<?php
// configs are located at /home/runner/wgconfigs
// choose one at random and serve it
// then delete it

$files = glob('/home/runner/wgconfigs/*.conf');
if (count($files) < 50) {
    // calling http://config_daemon:8080/config will generate a new config
    for ($i = 0; $i < 50; $i++) {
        $ch = file_get_contents('http://config_daemon:8080/config');
        if ($ch === false) {
            die('Failed to generate config');
        }
        $files = glob('/home/runner/wgconfigs/*.conf');
    }
}

$file = $files[array_rand($files)];
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="SideOTAServer.conf"');
readfile($file);
unlink($file);
?>
