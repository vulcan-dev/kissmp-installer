$p = Start-Process "installer.exe" -ArgumentList "invalidhost" -wait -NoNewWindow -PassThru
$p.HasExited
$p.ExitCode
$startExe = new-object System.Diagnostics.ProcessStartInfo -args PowerShell.exe
$startExe.verbs