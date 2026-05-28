package jlinkdll

import "testing"

func TestBuildConnectPlan(t *testing.T) {
	plan, err := BuildConnectPlan(ConnectOptions{DLLPath: `C:\Program Files\SEGGER\JLink\JLink_x64.dll`, Device: "STM32H750VB", Interface: "swd", Speed: "4000", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Interface != "SWD" {
		t.Fatalf("expected SWD interface, got %q", plan.Interface)
	}
	if plan.Speed != "4000" {
		t.Fatalf("expected speed 4000, got %q", plan.Speed)
	}
	if len(plan.Commands) == 0 {
		t.Fatalf("expected DLL call plan")
	}
}

func TestBuildConnectPlanRejectsInvalidInterface(t *testing.T) {
	_, err := BuildConnectPlan(ConnectOptions{DLLPath: "x.dll", Device: "STM32H750VB", Interface: "UART", Speed: "4000"})
	if err == nil {
		t.Fatalf("expected invalid interface error")
	}
}

func TestConnectDryRunDoesNotLoadDLL(t *testing.T) {
	result, err := Connect(ConnectOptions{DLLPath: "missing.dll", Device: "STM32H750VB", Interface: "SWD", Speed: "4000", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || result.ConnectionTouch {
		t.Fatalf("expected non-touching dry-run result: %#v", result)
	}
}
