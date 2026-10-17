$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms, System.Drawing
Add-Type @'
using System;
using System.Runtime.InteropServices;
[assembly: DefaultDllImportSearchPaths(DllImportSearchPath.System32)]
public class ElirCapture {
    [StructLayout(LayoutKind.Sequential)] public struct Point { public int X, Y; }
    [StructLayout(LayoutKind.Sequential)] public struct Rect { public int Left, Top, Right, Bottom; }
    [DllImport("user32.dll")] public static extern bool SetProcessDPIAware();
    [DllImport("user32.dll")] public static extern IntPtr WindowFromPoint(Point p);
    [DllImport("user32.dll")] public static extern IntPtr GetAncestor(IntPtr h, uint flags);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out Rect r);
}
'@
[void][ElirCapture]::SetProcessDPIAware()
$bounds = [System.Windows.Forms.SystemInformation]::VirtualScreen
$desktop = New-Object System.Drawing.Bitmap $bounds.Width, $bounds.Height
$graphics = [System.Drawing.Graphics]::FromImage($desktop)
$graphics.CopyFromScreen($bounds.Location, [System.Drawing.Point]::Empty, $bounds.Size)
$graphics.Dispose()
$form = New-Object System.Windows.Forms.Form
$form.FormBorderStyle = 'None'
$form.StartPosition = 'Manual'
$form.Bounds = $bounds
$form.TopMost = $true
$form.ShowInTaskbar = $false
$form.KeyPreview = $true
$form.Cursor = [System.Windows.Forms.Cursors]::Cross
$form.BackgroundImage = $desktop
$script:start = $null
$script:selection = [System.Drawing.Rectangle]::Empty
$script:accepted = $false
$form.Add_MouseDown({
    if ($_.Button -eq [System.Windows.Forms.MouseButtons]::Left) {
        $script:start = $_.Location
        $form.Capture = $true
    }
})
$form.Add_MouseMove({
    if ($null -ne $script:start) {
        $x = [Math]::Min($script:start.X, $_.X)
        $y = [Math]::Min($script:start.Y, $_.Y)
        $script:selection = New-Object System.Drawing.Rectangle $x, $y, ([Math]::Abs($_.X-$script:start.X)), ([Math]::Abs($_.Y-$script:start.Y))
        $form.Invalidate()
    }
})
$form.Add_Paint({
    if ($script:selection.Width -gt 0 -and $script:selection.Height -gt 0) {
        $_.Graphics.DrawRectangle([System.Drawing.Pens]::Red, $script:selection)
    }
})
$form.Add_MouseUp({
    if ($_.Button -eq [System.Windows.Forms.MouseButtons]::Left -and $null -ne $script:start) {
        $script:accepted = $script:selection.Width -gt 1 -and $script:selection.Height -gt 1
        $form.Close()
    }
})
$form.Add_KeyDown({
    if ($_.KeyCode -eq 'Escape') { $form.Close() }
    if ($_.KeyCode -eq 'Space') {
        $pos = [System.Windows.Forms.Cursor]::Position
        $form.Hide()
        [System.Windows.Forms.Application]::DoEvents()
        $point = New-Object ElirCapture+Point
        $point.X = $pos.X; $point.Y = $pos.Y
        $handle = [ElirCapture]::GetAncestor([ElirCapture]::WindowFromPoint($point), 2)
        $rect = New-Object ElirCapture+Rect
        if ([ElirCapture]::GetWindowRect($handle, [ref]$rect)) {
            $window = [System.Drawing.Rectangle]::FromLTRB($rect.Left, $rect.Top, $rect.Right, $rect.Bottom)
            $window = [System.Drawing.Rectangle]::Intersect($window, $bounds)
            $window.Offset(-$bounds.X, -$bounds.Y)
            $script:selection = $window
            $script:accepted = $window.Width -gt 0 -and $window.Height -gt 0
        }
        $form.Close()
    }
})
try {
    [void]$form.ShowDialog()
    if (-not $script:accepted) { exit 2 }
    $crop = $desktop.Clone($script:selection, $desktop.PixelFormat)
    try { $crop.Save($env:ELIR_SCREENSHOT_PATH, [System.Drawing.Imaging.ImageFormat]::Png) }
    finally { $crop.Dispose() }
} finally {
    $form.Dispose()
    $desktop.Dispose()
}
