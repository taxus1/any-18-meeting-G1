package acceptance

import "testing"

// 示例 spec（演示写法，业务线可删/替换）：默认只验证验收基座可编译可装配，
// 真正打接口/查库的用例在 StartApp() 之后写（需 DB_DSN/REDIS_ADDR，由 host_verify 注入）。
func TestSuiteWiring(t *testing.T) {
	s := NewSuite(t)
	if s.BaseURL == "" || s.User == "" || s.Pass == "" {
		t.Fatal("验收基座未就绪")
	}
}

// 真实用例骨架（取消 Skip 并起好 DB/Redis 后可用）：
//
//	func TestCreateItem(t *testing.T) {
//		s := NewSuite(t); defer s.Close()
//		if err := s.OpenDB(); err != nil { t.Skip("无 DB：", err) }
//		if err := s.StartApp(); err != nil { t.Fatal(err) }
//		name := fmt.Sprintf("it-%d", time.Now().UnixNano())
//		_, body := s.PostForm("/api/demo/item", map[string]string{"name": name, "score": "7"})
//		if !OK(body) { t.Fatalf("建项未探通：%s", body) }
//		if s.Count("t_demo_item", "name = ? and deleted_at is null", name) != 1 { t.Fatal("库里没落行") }
//	}
