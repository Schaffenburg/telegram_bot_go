TODO: rewrite zeit irgendwann; heute 0 uhr ist unschoen heute 23:59 schoener aber auch nicht gut :/

=> https://www.stwab.de/aschaffenburgGips/Gips?SessionMandant=Aschaffenburg&Anwendung=CMSWEBPAGE&Methode=RefreshHTMLAusgabe&RessourceID=710554&Container.Children:0.BezirkId=1648&Container.Children:0.Strasse=Dorfstra%26szlig%3Be
has ics download, so just parse that
CommandsDE/MuellPlanCommand.php:    protected $name = 'muellplan';
CommandsDE/MuellPlanCommand.php:    protected $description = 'Zeige den Abfallkalender von Wiesthal';
CommandsDE/MuellPlanCommand.php:    protected $usage = '/muellplan';
---
mit wuefel mby?
CommandsDE/WuerfelnCommand.php:    protected $name = 'wuerfeln';
---
CommandsDE/WerHatGeputztCommand.php:    protected $name = 'werhatgeputzt';
CommandsDE/WerHatGeputztCommand.php:    protected $description = 'Zeigt eine Zusammenfassung des Putzplans';
--
CommandsDE/IchHabGeputztCommand.php:    protected $name = 'ichhabgeputzt';
CommandsDE/IchHabGeputztCommand.php:    protected $description = 'Bestätigen, dass du den Space geputzt hast(Toilette)';
---
CommandsDE/WerBinIchCommand.php:    protected $name = 'werbinich';
CommandsDE/WerBinIchCommand.php:    protected $description = 'Zeige deine ID, Name und Username';
---
AUFWAND: MACHEN: TUN SEIN:
CommandsDE/SpracheCommand.php:    protected $name = 'sprache';
CommandsDE/SpracheCommand.php:    protected $description = 'Wähle in welcher Sprache der Bot mit dir reden soll';
--- find working goole scraper / api ---
CommandsDE/GidfCommand.php:    protected $name = 'gidf';
CommandsDE/GidfCommand.php:    protected $description = 'Google ist dein Freund';
--- find working api ---
CommandsDE/DatumCommand.php:    protected $name = 'datum';
CommandsDE/DatumCommand.php:    protected $description = 'Zeige Uhrzeit/Datum abhängig vom Standort';
--- fix ---
AliasCommands/StatusCommand.php:    protected $name = 'status'; 
---
Commands/ChangelogCommand.php:    protected $name = 'changelog';
--- check which sticker
Commands/AdminCommands/SendStickerCommand.php:    protected $name = 'sendsticker';
Commands/AdminCommands/SendStickerCommand.php:    protected $description = 'Let the bot post a certain sticker to Schaffen-CIX group';
---
Commands/IGoGetFoodCommand.php:    protected $name = 'igogetfood';
Commands/IWontComeCommand.php:    protected $name = 'iwontcome';
Commands/MustHaveCommand.php:    protected $name = 'musthave';
Commands/PingCommand.php:    protected $name = 'ping';
Commands/RollCommand.php:    protected $name = 'roll';
Commands/SlapCommand.php:    protected $name = 'slap';
Commands/UnsubscribeCommand.php:    protected $name = 'unsubscribe';
Commands/WhatistheanswerforCommand.php:    protected $name = 'whatistheanswerfor';
Commands/WhoamiCommand.php:    protected $name = 'whoami';
Commands/WhoHasCleanedCommand.php:    protected $name = 'whohascleaned';
Commands/WhoGetsFoodCommand.php:    protected $name = 'whogetsfood';
Commands/WhoIsThereCommand.php:    protected $name = 'whoisthere';
Commands/TimerCommand.php:    protected $name = 'timer';
Commands/SubscribeCommand.php:    protected $name = 'subscribe';
Commands/SystemCommands/DailyAutoCloseCommand.php:    protected $name = 'dailyautoclose';
Commands/SystemCommands/TrashReminderCommand.php:    protected $name = 'trashreminder';
Commands/SystemCommands/SQLMaintenanceCommand.php:    protected $name = 'sqlmaintenance';
AliasCommandsDE/PutzplanCommand.php:    protected $name = 'putzplan';
AliasCommandsDE/KommenCommand.php:    protected $name = 'kommen';
Commands/OpenSpaceCommand.php:    protected $name = 'openspace';
Commands/LolCommand.php:    protected $name = 'lol';
Commands/LanguageCommand.php:    protected $name = 'language';
Commands/IComeTodayCommand.php:    protected $name = 'icometoday';
Commands/ILeaveNowCommand.php:    protected $name = 'ileavenow';
Commands/ICleanedCommand.php:    protected $name = 'icleaned';
AliasCommandsDE/IchHolHeuteEssenCommand.php:    protected $name = 'ichholheuteessen';
Commands/HowHotIsItCommand.php:    protected $name = 'howhotisit';
Commands/HeaterOffCommand.php:    protected $name = 'heateroff';
Commands/HeaterOnCommand.php:    protected $name = 'heateron';
AliasCommandsDE/IchHoleEssenCommand.php:    protected $name = 'ichholeessen';
AliasCommandsDE/IchHabeGeputztCommand.php:    protected $name = 'ichhabegeputzt';
AliasCommandsDE/HabGeputzt.php:    protected $name = 'habgeputzt';
Commands/DateCommand.php:    protected $name = 'date';
Commands/CloseSpaceCommand.php:    protected $name = 'closespace';
AliasCommands/IAmThereCommand.php:    protected $name = 'iamthere';
