<?php
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="SideOTAServer.conf"');
file_get_contents('http://config_daemon:8080/config');
?>
