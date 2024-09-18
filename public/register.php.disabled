<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SideOTAJIT</title>
    <link rel="stylesheet" href="main.css">
</head>

<body>
    <nav class="head">
        <a class="noborder button" href="index.php">SideOTAJIT</a>
        <div>
            <a class="button" href="register.php">Register</a>
            <a class="button" href="activate.php">Activate</a>
        </div>
    </nav>
    <?php
    if ($_SERVER['REQUEST_METHOD'] === 'POST') {
        if (isset($_FILES['pairing']) && $_FILES['pairing']['error'] === UPLOAD_ERR_OK) {
            $file = $_FILES['pairing'];
            $fileTmpPath = $file['tmp_name'];
            $fileName = $file['name'];
            $fileSize = $file['size'];
            $fileType = $file['type'];

            // Check the file contents
            $fileContents = file_get_contents($fileTmpPath);

            $fileHash = hash_file('blake2b', $fileTmpPath);

            $uploadDir = '/home/runner/.pymobiledevice';

            // Create the directory if it doesn't exist
            if (!is_dir($uploadDir)) {
                mkdir($uploadDir, 0755, true);
            }

            $destinationPath = $uploadDir . $fileHash . '.plist';

            if ($fileSize > 1024 * 1024) {
                die('File too large');
            }

            // Move the file to the destination
            move_uploaded_file($fileTmpPath, $destinationPath);

            echo 'File uploaded successfully';
        } else {
            echo 'Error uploading file';
        }
    }
    ?>

    <form action="register.php" method="post">
        <h1 class="text-center">Register</h1>
        <div class="everything-center">
            <div>
                <h3 class="dblock text-center">Pairing:</h3>
                <div class="everything-center">
                    <input type="file" name="pairing" id="pairing">
                </div>
                <br>
                <!-- we had to cheap out on the UI, sorry -->
                <div class="everything-center">
                    <input type="submit" value="Register">
                </div>
            </div>
        </div>
    </form>
</body>

</html>
