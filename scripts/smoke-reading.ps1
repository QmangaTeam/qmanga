param([Parameter(Mandatory=$true)][string]$ReleaseId,[string]$AppUrl='http://localhost:8080',[string]$MailboxUrl='http://localhost:8025')
$ErrorActionPreference='Stop'
if (-not ([uri]$AppUrl).IsLoopback -or -not ([uri]$MailboxUrl).IsLoopback) { throw 'Local test only.' }
$clients=[Collections.Generic.List[System.Net.Http.HttpClient]]::new()
function New-Client {
 $handler=[System.Net.Http.HttpClientHandler]::new();$handler.CookieContainer=[Net.CookieContainer]::new();$client=[System.Net.Http.HttpClient]::new($handler);$client.Timeout=[TimeSpan]::FromSeconds(20);$clients.Add($client);return $client
}
function Send($Client,[string]$Method,[string]$Path,$Body=$null,[string]$Csrf='',[int]$Expected=200){
 $request=[System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::new($Method),"$AppUrl/api/v1/$Path");$request.Headers.Add('Origin',([uri]$AppUrl).GetLeftPart([UriPartial]::Authority));if($Csrf){$request.Headers.Add('X-CSRF-Token',$Csrf)}
 if($null -ne $Body){$request.Content=[System.Net.Http.StringContent]::new(($Body|ConvertTo-Json -Depth 10 -Compress),[Text.Encoding]::UTF8,'application/json')}
 try{$response=$Client.SendAsync($request).GetAwaiter().GetResult();try{$text=$response.Content.ReadAsStringAsync().GetAwaiter().GetResult();$value=if($text){$text|ConvertFrom-Json}else{$null};if([int]$response.StatusCode -ne $Expected){throw "$Method $Path expected $Expected, got $([int]$response.StatusCode) $($value.error.code)"};return $value}finally{$response.Dispose()}}finally{$request.Dispose()}
}
function Token([string]$Email){
 for($attempt=0;$attempt -lt 30;$attempt++){$mail=Invoke-RestMethod -TimeoutSec 5 "$MailboxUrl/api/v1/messages";$found=@($mail.messages|Where-Object{@($_.To|Where-Object{$_.Address -eq $Email}).Count -gt 0});if($found.Count){$message=Invoke-RestMethod -TimeoutSec 5 "$MailboxUrl/api/v1/message/$($found[0].ID)";if($message.Text -match '/verify-email\?token=([A-Za-z0-9_-]{43})'){return $Matches[1]}};Start-Sleep -Milliseconds 200};throw 'Verification mail missing.'
}
function Account([bool]$Admin=$false){
 $client=New-Client;$email="reading-$([guid]::NewGuid().ToString('N'))@example.test";$password=[guid]::NewGuid().ToString('N')+'-Aa1'
 if($Admin){$password|docker compose exec -T -e "ADMIN_EMAIL=$email" api qmanga-admin|Out-Null;if($LASTEXITCODE -ne 0){throw 'Admin setup failed.'}}else{$null=Send $client POST auth/register @{email=$email;password=$password} -Expected 202;$null=Send $client POST auth/verify-email @{token=(Token $email)}}
 $null=Send $client POST auth/login @{email=$email;password=$password};$user=(Send $client GET auth/me).user;$csrf=(Send $client GET auth/csrf).csrf_token;$client.DefaultRequestHeaders.Add('X-Qmanga-Account',$user.id);return @{Client=$client;Id=$user.id;Csrf=$csrf;Email=$email}
}
try{
 $anonymous=New-Client;$manifest=Send $anonymous GET "releases/$ReleaseId/manifest";$title=$manifest.title_id;$page=$manifest.pages[0].id
 $baseline=Send $anonymous GET "community/titles/$title`?release=$ReleaseId"
 $first=Account;$second=Account;$moderator=Account $true
 $null=Send $first.Client PUT "library/$title" @{status='reading';favorite=$true;preferred_team_id=$manifest.team.id;notify=$true} $first.Csrf 204
 $library=Send $first.Client GET library;if(@($library.items).Count -lt 1){throw 'Library empty.'};if(@((Send $second.Client GET library).items).Count -ne 0){throw 'Cross-account library leak.'}
 $event=@{event_id=[guid]::NewGuid().ToString();device_id='smoke-phone';sequence=1;base_version=0;release_id=$ReleaseId;revision_id=$manifest.revision_id;page_id=$page;fraction=0.3;completed=$false}
 $result=Send $first.Client POST "progress/$title" $event $first.Csrf;if($result.progress.version -ne 1){throw 'Progress not versioned.'}
 if((Send $first.Client POST "progress/$title" $event $first.Csrf).outcome -ne 'duplicate'){throw 'Replay was not idempotent.'}
 $event.event_id=[guid]::NewGuid().ToString();$event.device_id='smoke-tablet';$result=Send $first.Client POST "progress/$title" $event $first.Csrf 409;if($result.outcome -ne 'conflict'){throw 'Stale device overwrote position.'}
 $event.base_version=1;$null=Send $first.Client POST "progress/$title" $event $first.Csrf
 $second.Client.DefaultRequestHeaders.Remove('X-Qmanga-Account')|Out-Null;$second.Client.DefaultRequestHeaders.Add('X-Qmanga-Account',$first.Id);$null=Send $second.Client GET "progress/$title" -Expected 409;$second.Client.DefaultRequestHeaders.Remove('X-Qmanga-Account')|Out-Null;$second.Client.DefaultRequestHeaders.Add('X-Qmanga-Account',$second.Id)
 $comment=Send $first.Client POST community/comments @{title_id=$title;release_id=$ReleaseId;parent_id='';body='Жергілікті тексеру пікірі';spoiler=$true} $first.Csrf 201
 $reply=Send $second.Client POST community/comments @{title_id=$title;release_id=$ReleaseId;parent_id=$comment.id;body='Жергілікті тексеру жауабы';spoiler=$false} $second.Csrf 201
 $null=Send $second.Client PUT "community/comments/$($comment.id)" @{body='Өзгенің мәтінін өзгерту';spoiler=$false} $second.Csrf 403
 $null=Send $first.Client POST community/comments @{title_id=$title;release_id=$ReleaseId;parent_id=$reply.id;body='Терең жауап қабылданбайды';spoiler=$false} $first.Csrf 422
 $null=Send $first.Client PUT "community/titles/$title/rating" @{score=9} $first.Csrf 204;$null=Send $first.Client PUT "community/titles/$title/rating" @{score=8} $first.Csrf 204
 $discussion=Send $anonymous GET "community/titles/$title`?release=$ReleaseId";$expectedScore=($baseline.score*$baseline.ratings+8)/($baseline.ratings+1);if($discussion.ratings -ne ($baseline.ratings+1) -or [Math]::Abs($discussion.score-$expectedScore) -gt 0.00001){throw 'Rating was duplicated.'}
 $report=Send $second.Client POST community/reports @{title_id=$title;release_id=$ReleaseId;page_id=$page;target_type='page';target_id=$page;category='image';reason='Жергілікті сурет тексеруі'} $second.Csrf 201
 $null=Send $second.Client GET moderation/reports -Expected 403
 $null=Send $moderator.Client PUT "moderation/reports/$($report.id)" @{status='resolved';reason='Жергілікті тексеру аяқталды'} $moderator.Csrf 204
 $notified=$false;for($attempt=0;$attempt -lt 20;$attempt++){$notifications=Send $first.Client GET notifications;if(@($notifications.items|Where-Object{$_.kind -eq 'comment_reply'}).Count){$notified=$true;break};Start-Sleep -Milliseconds 500};if(-not $notified){throw 'Reply notification missing.'}
 $null=Send $anonymous GET notifications -Expected 401
 $html=$first.Client.GetStringAsync("$AppUrl/read/$ReleaseId").GetAwaiter().GetResult();if($html.Contains($first.Email)){throw 'Private account data in public reader HTML.'}
 $null=Send $first.Client POST auth/logout @{} $first.Csrf 204;$null=Send $first.Client GET "progress/$title" -Expected 401
 Write-Output 'Reading smoke passed: library isolation, progress replay/conflicts/account guard, ratings, comments, reports, worker notifications and reader privacy.'
}finally{foreach($client in $clients){$client.Dispose()}}
