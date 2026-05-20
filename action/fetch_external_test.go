package action

import (
	"proxy-provider/core"
	"testing"
)

func TestFetchExternalRejectsExecutionWhenDisabled(t *testing.T) {
	ctx := core.NewExecuteContext()

	err := (&FetchExternalAction{Command: "unused"}).Execute(ctx)
	if err == nil {
		t.Fatal("expected disabled fetch-external to fail")
	}
	if want := "fetch-external is disabled, start proxy-formatter with -allow-external-script to enable it"; err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}
