package initializer

import "testing"

func TestRedisInit(t *testing.T) {
	prev := RDB
	t.Cleanup(func() {
		if RDB != nil && RDB != prev {
			_ = RDB.Close()
		}
		RDB = prev
	})

	t.Setenv("REDIS_HOST", "localhost:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DATABASE", "2")

	RDB = nil
	RedisInit()
	if RDB == nil {
		t.Fatal("RDB is nil after RedisInit")
	}
	opts := RDB.Options()
	if opts.Addr != "localhost:6379" {
		t.Errorf("Addr = %q", opts.Addr)
	}
	if opts.Password != "secret" {
		t.Errorf("Password = %q", opts.Password)
	}
	if opts.DB != 2 {
		t.Errorf("DB = %d, want 2", opts.DB)
	}
}

func TestRedisInitInvalidDatabase(t *testing.T) {
	prev := RDB
	t.Cleanup(func() {
		if RDB != nil && RDB != prev {
			_ = RDB.Close()
		}
		RDB = prev
	})

	t.Setenv("REDIS_HOST", "localhost:6379")
	t.Setenv("REDIS_DATABASE", "not-a-number")

	RedisInit()
	if RDB == nil {
		t.Fatal("RDB is nil after RedisInit")
	}
	if RDB.Options().DB != 0 {
		t.Errorf("DB = %d, want 0 on invalid input", RDB.Options().DB)
	}
}
