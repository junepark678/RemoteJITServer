<?php
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="SideOTAServer.conf"');
echo file_get_contents('http://10.9.0.1:6969/config');
?>
