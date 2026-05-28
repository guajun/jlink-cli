package jlinkdll

var RequiredSymbols = []string{
	"JLINKARM_OpenEx",
	"JLINKARM_Close",
	"JLINKARM_ExecCommand",
	"JLINKARM_Connect",
	"JLINKARM_IsOpen",
	"JLINKARM_EMU_IsConnected",
	"JLINKARM_IsConnected",
	"JLINKARM_TIF_Select",
	"JLINKARM_SetSpeed",
	"JLINKARM_Halt",
	"JLINKARM_Go",
	"JLINKARM_IsHalted",
	"JLINKARM_ReadMemEx",
	"JLINKARM_ReadReg",
	"JLINKARM_ReadRegs",
	"JLINKARM_SetBPEx",
	"JLINKARM_ClrBPEx",
	"JLINK_DownloadFile",
	"JLINKARM_BeginDownload",
	"JLINKARM_WriteMem",
	"JLINKARM_EndDownload",
}

func Doctor(opts Options) DoctorReport {
	env := opts.Env
	report := DoctorReport{
		Candidates:        CandidatePaths(opts),
		Platform:          platform(),
		ConnectionTouched: false,
	}
	if env == nil {
		env = environ()
	}
	report.Environment = map[string]string{
		"JLINK_DLL_PATH": getenv(env, "JLINK_DLL_PATH"),
		"JLINK_PATH":     getenv(env, "JLINK_PATH"),
	}
	for index := range report.Candidates {
		if !report.Candidates[index].Exists {
			continue
		}
		symbols, err := ProbeSymbols(report.Candidates[index].Path, RequiredSymbols)
		report.Candidates[index].LoadOK = err == nil
		if err != nil {
			report.Candidates[index].Reason = err.Error()
			continue
		}
		report.Selected = report.Candidates[index].Path
		report.RuntimeLoad = true
		report.Symbols = symbols
		for _, symbol := range symbols {
			if !symbol.Present {
				report.MissingRequired = append(report.MissingRequired, symbol.Name)
			}
		}
		return report
	}
	return report
}
