# Powershell doesn't default to UTF8 just yet, so we're forcing it as there are too many problems
# that pop up when we don't
[console]::InputEncoding = [console]::OutputEncoding = New-Object System.Text.UTF8Encoding
$env:POWERLINE_COMMAND = "oh-my-posh"
$env:CONDA_PROMPT_MODIFIER = $false

# specific module support (disabled by default)
function Set-DefaultEnvValue {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = $true)]
        [string]
        $Name
    )

    $value = [System.Environment]::GetEnvironmentVariable($Name)
    if ($value -eq $null) {
        [System.Environment]::SetEnvironmentVariable($Name, $false)
    }
}
Set-DefaultEnvValue("AZ_ENABLED")
Set-DefaultEnvValue("POSH_GIT_ENABLED")

$global:PoshSettings = New-Object -TypeName PSObject -Property @{
    Theme = "";
    EnableToolTips = $false;
}

# used to detect empty hit
$global:omp_lastHistoryId = -1

$config = "::CONFIG::"
if (Test-Path $config) {
    $global:PoshSettings.Theme = (Resolve-Path -Path $config).ProviderPath
}

function global:Set-PoshContext {}

function global:Initialize-ModuleSupport {
    if ($env:POSH_GIT_ENABLED -eq $true -and (Get-Module -Name "posh-git")) {
        [Diagnostics.CodeAnalysis.SuppressMessageAttribute('PSProvideCommentHelp', '', Justification = 'Variable used later(not in this scope)')]
        $global:GitStatus = Get-GitStatus
        $env:POSH_GIT_STATUS = Write-GitStatus -Status $global:GitStatus
    }

    $env:AZ_SUBSCRIPTION_NAME = $null
    $env:AZ_SUBSCRIPTION_ID = $null

    if ($env:AZ_ENABLED -eq $true) {
        try {
            $subscription = Get-AzContext | Select-Object -ExpandProperty "Subscription" | Select-Object "Name", "Id", "Account"
            if ($null -ne $subscription) {
                $env:AZ_SUBSCRIPTION_NAME = $subscription.Name
                $env:AZ_SUBSCRIPTION_ID = $subscription.Id
                $env:AZ_SUBSCRIPTION_ACCOUNT = $subscription.Account
            }
        }
        catch {}
    }

    # Set the keyhandler to enable tooltips
    if ($global:PoshSettings.EnableToolTips -eq $true) {
        Set-PSReadlineKeyHandler -Key SpaceBar -ScriptBlock {
            [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
            $position = $host.UI.RawUI.CursorPosition
            $omp = "::OMP::"
            $config = $global:PoshSettings.Theme
            $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
            $cleanPSWD = $PWD.ToString().TrimEnd("\")
            $tooltip = $null
            $cursor = $null
            [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$tooltip, [ref]$cursor)
            $standardOut = @(&$omp --pwd="$cleanPWD" --pswd="$cleanPSWD" --config="$config" --tooltip="$tooltip" 2>&1)
            Write-Host $standardOut -NoNewline
            $host.UI.RawUI.CursorPosition = $position
        }
    }
}

[ScriptBlock]$Prompt = {
    #store if the last command was successful
    $lastCommandSuccess = $?
    #store the last exit code for restore
    $realLASTEXITCODE = $global:LASTEXITCODE
    $errorCode = 0
    Initialize-ModuleSupport
    Set-PoshContext
    if ($lastCommandSuccess -eq $false) {
        #native app exit code
        if ($realLASTEXITCODE -is [int] -and $realLASTEXITCODE -gt 0) {
            $errorCode = $realLASTEXITCODE
        }
        else {
            $errorCode = 1
        }
    }

    # read stack count from current stack(if invoked from profile=right value,otherwise use the global variable set in Set-PoshPrompt(stack scoped to module))
    $stackCount = (Get-Location -stack).Count
    try {
        if ($global:omp_global_sessionstate -ne $null) {
            $stackCount = ($global:omp_global_sessionstate).path.locationstack('').count
        }
    }
    catch {}

    $executionTime = -1
    $history = Get-History -ErrorAction Ignore -Count 1
    if ($null -ne $history -and $null -ne $history.EndExecutionTime -and $null -ne $history.StartExecutionTime -and $global:omp_lastHistoryId -ne $history.Id) {
        $executionTime = ($history.EndExecutionTime - $history.StartExecutionTime).TotalMilliseconds
        $global:omp_lastHistoryId = $history.Id
    }
    $omp = "::OMP::"
    $config = $global:PoshSettings.Theme
    $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
    $cleanPSWD = $PWD.ToString().TrimEnd("\")
    $standardOut = @(&$omp --error="$errorCode" --pwd="$cleanPWD" --pswd="$cleanPSWD" --execution-time="$executionTime" --stack-count="$stackCount" --config="$config" 2>&1)
    # the output can be multiline, joining these ensures proper rendering by adding line breaks with `n
    $standardOut -join "`n"
    $global:LASTEXITCODE = $realLASTEXITCODE
    #remove temp variables
    Remove-Variable realLASTEXITCODE -Confirm:$false
    Remove-Variable lastCommandSuccess -Confirm:$false
}
Set-Item -Path Function:prompt -Value $Prompt -Force

function global:Write-PoshDebug {
    $omp = "::OMP::"
    $config = $global:PoshSettings.Theme
    $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
    $cleanPSWD = $PWD.ToString().TrimEnd("\")
    $standardOut = @(&$omp --error=1337 --pwd="$cleanPWD" --pswd="$cleanPSWD" --execution-time=9001 --config="$config" --debug 2>&1)
    $standardOut -join "`n"
}

<#
.SYNOPSIS
    Exports the current oh-my-posh theme
.DESCRIPTION
    By default the config is exported in json to the clipboard
.EXAMPLE
    Export-PoshTheme
    Current theme exported in json to clipboard
.EXAMPLE
    Export-PoshTheme -Format toml
    Current theme exported in toml to clipboard
.EXAMPLE
    Export-PoshTheme c:\temp\theme.toml toml
    Current theme exported in toml to c:\temp\theme.toml
.EXAMPLE
    Export-PoshTheme ~\theme.toml toml
    Current theme exported in toml to your home\theme.toml
#>
function global:Export-PoshTheme {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        # The file path where the theme will be exported. If not provided, the config is copied to the clipboard by default.
        $FilePath,
        [Parameter(Mandatory = $false)]
        [ValidateSet('json', 'yaml', 'toml')]
        [string]
        # The format of the theme
        $Format = 'json'
    )

    $config = $global:PoshSettings.Theme
    $omp = "::OMP::"
    $configString = @(&$omp --config="$config" --config-format="$Format" --print-config 2>&1)
    # if no path, copy to clipboard by default
    if ($FilePath -ne "") {
        #https://stackoverflow.com/questions/3038337/powershell-resolve-path-that-might-not-exist
        $FilePath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($FilePath)
        [IO.File]::WriteAllLines($FilePath, $configString)
    }
    else {
        Set-Clipboard $configString
        Write-Output "Theme copied to clipboard"
    }
}

function global:Export-PoshImage {
    param(
        [Parameter(Mandatory = $false)]
        [int]
        $RPromptOffset = 40,
        [Parameter(Mandatory = $false)]
        [int]
        $CursorPadding = 30,
        [Parameter(Mandatory = $false)]
        [string]
        $Author
    )

    if ($Author) {
        $Author = "--author=$Author"
    }

    $omp = "::OMP::"
    $config = $global:PoshSettings.Theme
    $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
    $cleanPSWD = $PWD.ToString().TrimEnd("\")
    $standardOut = @(&$omp --config="$config" --pwd="$cleanPWD" --pswd="$cleanPSWD" --export-png --rprompt-offset="$RPromptOffset" --cursor-padding="$CursorPadding" $Author 2>&1)
    $standardOut -join "`n"
}
