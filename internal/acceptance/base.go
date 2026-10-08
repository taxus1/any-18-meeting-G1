package acceptance

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Suite：Go/Spock 对位的验收基座（Go 版）。
// 契约无关：不 import 业务包，只「编译并起产物进程 → 打 HTTP / 查 DB」。
// 业务线 spec 放 cases/<迭代>/（package acceptance），继承/复用本文件（放 cases/common/）。
type Suite struct {
	T       *testing.T
	BaseURL string
	User    string
	Pass    string
	DB      *sql.DB
	proc    *exec.Cmd
	client  *http.Client
	bin     string
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func NewSuite(t *testing.T) *Suite {
	return &Suite{
		T:       t,
		BaseURL: env("APP_BASE_URL", "http://127.0.0.1:18080"),
		User:    env("APP_USER", "admin"),
		Pass:    env("APP_PASS", "admin123"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// StartApp 编译并启动产物进程（契约无关：不依赖具体包/路由名）。
// 注意：go test 的 CWD 是**包目录**（internal/acceptance），不是模块根 —— 必须回到模块根再 build。
func (s *Suite) StartApp() error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	// 注意：镜像里 /tmp 常挂 noexec → 编到模块目录下再执行（Close 时清理）。
	bin := filepath.Join(root, ".somepro-app")
	s.bin = bin
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = root
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("go build 失败：%w", err)
	}
	cmd := exec.Command(bin)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "SERVER_PORT="+portOf(s.BaseURL))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	s.proc = cmd
	return s.waitReady(30 * time.Second)
}

// moduleRoot：从 CWD 向上找到含 go.mod 的目录（go test 的 CWD 是包目录）。
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("向上找不到 go.mod")
		}
		dir = parent
	}
}

// portOf：从 "http://127.0.0.1:18080" 取端口（SERVER_PORT 只能是端口，不能是 host:port）。
func portOf(u string) string {
	u = strings.TrimPrefix(strings.TrimPrefix(u, "http://"), "https://")
	if i := strings.LastIndex(u, ":"); i >= 0 {
		return u[i+1:]
	}
	return "8080"
}

func (s *Suite) waitReady(d time.Duration) error {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		resp, err := s.do("GET", "/", nil)
		if err == nil {
			_ = resp
			return nil // 端口通了即可（路由可能 401/404，说明服务已起）
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("服务在 %s 内没起来", d)
}

func (s *Suite) Close() {
	if s.proc != nil && s.proc.Process != nil {
		_ = s.proc.Process.Kill()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}
	if s.bin != "" {
		_ = os.Remove(s.bin)
	}
}

// ---------------- HTTP ----------------

func (s *Suite) do(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, s.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.User, s.Pass)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return s.client.Do(req)
}

func (s *Suite) Get(path string) (int, string) {
	resp, err := s.do("GET", path, nil)
	if err != nil {
		return -1, ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func (s *Suite) PostForm(path string, form map[string]string) (int, string) {
	var buf bytes.Buffer
	first := true
	for k, v := range form {
		if !first {
			buf.WriteByte('&')
		}
		first = false
		buf.WriteString(url.QueryEscape(k) + "=" + url.QueryEscape(v))
	}
	resp, err := s.do("POST", path, &buf)
	if err != nil {
		return -1, ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func (s *Suite) PostJSON(path, body string) (int, string) {
	req, err := http.NewRequest("POST", s.BaseURL+path, strings.NewReader(body))
	if err != nil {
		return -1, ""
	}
	req.SetBasicAuth(s.User, s.Pass)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return -1, ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

// OK：成败看 Result.code=0（不看 HTTP 状态码 —— 脚手架失败也回 200）。
func OK(body string) bool {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		return false
	}
	c, ok := m["code"].(float64)
	return ok && c == 0
}

// ---------------- DB ----------------

func (s *Suite) OpenDB() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return fmt.Errorf("缺 DB_DSN")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	s.DB = db
	return db.Ping()
}

func (s *Suite) Count(table, where string, args ...interface{}) int64 {
	var n int64
	q := "select count(*) from " + table
	if where != "" {
		q += " where " + where
	}
	if err := s.DB.QueryRow(q, args...).Scan(&n); err != nil {
		s.T.Fatalf("count 失败：%v", err)
	}
	return n
}

func (s *Suite) Exec(sqlText string, args ...interface{}) {
	if _, err := s.DB.Exec(sqlText, args...); err != nil {
		s.T.Fatalf("exec 失败：%v", err)
	}
}
