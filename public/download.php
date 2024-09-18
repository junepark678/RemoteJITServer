<?php
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="SideOTAServer.conf"');
echo file_get_contents('http://config_daemon:6969/config');
?>
