# MVP 測試用 SMB 目錄

在 Windows PowerShell 執行 `scripts/setup-test-smb.ps1`，腳本會複製系統的
Notepad 作為三個可安全啟動的 portable 測試程式。接著將設定檔的
`smbBasePath` 指向此目錄，即可驗證下載、更新、修復、解除安裝與離線啟動。

測試完成後可刪除三個腳本產生的 `.exe`；它們不應提交到版本庫。
