$env:PATH = 'C:\Program Files\Go\bin;C:\msys64\mingw64\bin;' + $env:PATH
$env:CGO_ENABLED = '1'
$env:GOARCH = 'amd64'
$env:GOOS = 'windows'
Set-Location C:\SetTracker

# ── Generar bear.ico + rsrc.syso si no existe o si bear.png cambió ──
$needIcon = (-not (Test-Path 'rsrc.syso')) -or
            ((Get-Item 'assets\bear.png').LastWriteTime -gt (Get-Item 'rsrc.syso').LastWriteTime)

if ($needIcon) {
    Write-Output "Generando recurso de icono..."

    # 1. Convertir bear.png a ICO multi-tamaño (PNG-in-ICO, Vista+)
    Add-Type -AssemblyName System.Drawing
    $sizes = @(256, 128, 64, 48, 32, 16)
    $src   = New-Object System.Drawing.Bitmap("$PWD\assets\bear.png")
    $pngImages = @{}

    foreach ($s in $sizes) {
        $bmp = New-Object System.Drawing.Bitmap($s, $s,
            ([System.Drawing.Imaging.PixelFormat]::Format32bppArgb))
        $g = [System.Drawing.Graphics]::FromImage($bmp)
        $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $g.DrawImage($src, 0, 0, $s, $s)
        $g.Dispose()
        $ms = New-Object System.IO.MemoryStream
        $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
        $pngImages[$s] = $ms.ToArray()
        $ms.Dispose()
        $bmp.Dispose()
    }
    $src.Dispose()

    # Estructura ICO: header (6b) + directorio (16b × N) + datos PNG
    $out = New-Object System.IO.MemoryStream
    $bw  = New-Object System.IO.BinaryWriter($out)
    $bw.Write([uint16]0)              # reservado
    $bw.Write([uint16]1)              # tipo = icono
    $bw.Write([uint16]$sizes.Count)   # cantidad

    $offset = 6 + 16 * $sizes.Count
    foreach ($s in $sizes) {
        $d  = $pngImages[$s]
        $ds = if ($s -ge 256) { 0 } else { $s }   # 0 = 256 en ICO
        $bw.Write([byte]$ds);   $bw.Write([byte]$ds)
        $bw.Write([byte]0);     $bw.Write([byte]0)    # colores, reservado
        $bw.Write([uint16]1);   $bw.Write([uint16]32) # planes, bpp
        $bw.Write([uint32]$d.Length)
        $bw.Write([uint32]$offset)
        $offset += $d.Length
    }
    foreach ($s in $sizes) { $bw.Write($pngImages[$s]) }
    $bw.Flush()
    [System.IO.File]::WriteAllBytes("$PWD\assets\bear.ico", $out.ToArray())
    $bw.Dispose(); $out.Dispose()
    Write-Output "  bear.ico creado ($($sizes.Count) tamaños)"

    # 2. Crear script de recurso Windows
    Set-Content -Path "$PWD\assets\icon.rc" -Value 'IDI_ICON1 ICON "bear.ico"'

    # 3. Compilar con windres (MinGW)
    # rsrc.syso debe quedar en la raíz para que el linker de Go lo encuentre
    $wr = Start-Process -FilePath 'C:\msys64\mingw64\bin\windres.exe' `
        -ArgumentList '-i', 'assets\icon.rc', '-o', 'rsrc.syso', '-O', 'coff' `
        -Wait -PassThru -NoNewWindow `
        -RedirectStandardError "$PWD\windres_err.txt"

    if ($wr.ExitCode -ne 0) {
        Write-Output "ERROR windres:"
        Get-Content "$PWD\windres_err.txt"
        exit 1
    }
    Write-Output "  rsrc.syso compilado"
}

# ── Go build ──
$proc = Start-Process -FilePath 'cmd.exe' `
    -ArgumentList '/c', 'go build -ldflags="-H windowsgui -s -w" -o BearTracker.exe . > build_err.txt 2>&1 && echo 0 > build_exit.txt || echo 1 > build_exit.txt' `
    -Wait -PassThru -NoNewWindow
