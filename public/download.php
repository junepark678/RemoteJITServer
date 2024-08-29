<?php
// configs are located at /home/runner/wgconfigs
// choose one at random and serve it
// then delete it

$files = glob('/home/runner/wgconfigs/*.conf');
if (count($files) === 0) {
    die('No configs available');
}

$file = $files[array_rand($files)];
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="SideOTAServer.conf"');
readfile($file);
unlink($file);
?>
