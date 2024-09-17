<?php
// configs are located at /home/runner/wgconfigs
// choose one at random and serve it
// then delete it

$files = glob('/home/runner/wgconfigs/*');

header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="SideOTAServer.conf"');
file_get_contents('http://config_daemon:8080/config');
?>
