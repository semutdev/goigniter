<!DOCTYPE html>
<html>
<head>
    <title><?= $title ?></title>
    <link rel="stylesheet" href="<?= base_url('css/styles.css') ?>">
</head>
<body>
    <div class="user-card">
        <h1><?= htmlspecialchars($title) ?>: <?= htmlspecialchars($user->name) ?></h1>
        <div class="user-info">
            <p><strong>ID:</strong> <?= $user->id ?></p>
            <p><strong>Email:</strong> <?= $user->email ?></p>
            <p><strong>Status:</strong>
                <?php if ($user->is_active): ?>
                    <span class="badge active">Active</span>
                <?php else: ?>
                    <span class="badge inactive">Inactive</span>
                <?php endif; ?>
            </p>
        </div>
        <p><a href="<?= base_url('users') ?>">Back to List</a></p>
    </div>
</body>
</html>
