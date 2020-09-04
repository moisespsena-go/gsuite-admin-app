package gsuite_admin_app

import "context"

func Context(ctx ...context.Context) (Ctx context.Context) {
	for _, ctx := range ctx {
		if ctx != nil {
			Ctx = ctx
		}
	}
	if Ctx == nil {
		Ctx = context.Background()
	}
	return
}
