<!DOCTYPE html>
<html>
<head>
    <title><?= $title ?></title>
    <link rel="stylesheet" href="<?= base_url('css/styles.css') ?>">
</head>
<body>
    <div class="container">
        <h1><?= $title ?></h1>
        <p><a href="<?= base_url('users/create') ?>">Create User</a></p>

        <?php if (!empty($users)): ?>
        <table class="table">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                <?php foreach ($users as $user): ?>
                <tr>
                    <td><?= $user->id ?></td>
                    <td><?= htmlspecialchars($user->name) ?></td>
                    <td><?= $user->email ?></td>
                    <td>
                        <?php if ($user->is_active): ?>
                            <span class="badge active">Active</span>
                        <?php else: ?>
                            <span class="badge inactive">Inactive</span>
                        <?php endif; ?>
                    </td>
                    <td>
                        <a href="<?= site_url('users/detail/' . $user->id) ?>">View</a>
                    </td>
                </tr>
                <?php endforeach; ?>
            </tbody>
        </table>
        <?php else: ?>
            <p>No users found.</p>
        <?php endif; ?>
    </div>
</body>
</html>
