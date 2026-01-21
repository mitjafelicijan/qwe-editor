<?php

// A simple PHP script

namespace App;

use Exception;

class User {
    private $name;
    private $email;

    public function __construct($name, $email) {
        $this->name = $name;
        $this->email = $email;
    }

    public function getName() {
        return $this->name;
    }

    public function getEmail() {
        return $this->email;
    }

    public function save() {
        // Simulate saving to database
        echo "Saving user {$this->name} to database.\n";
        return true;
    }
}

function processUsers($users) {
    foreach ($users as $user) {
        try {
            if ($user->save()) {
                echo "User saved successfully.\n";
            }
        } catch (Exception $e) {
            echo "Error: " . $e->getMessage() . "\n";
        }
    }
}

// Data array
$users_data = [
    ['name' => 'Alice', 'email' => 'alice@example.com'],
    ['name' => 'Bob', 'email' => 'bob@example.com'],
];

$user_objects = [];

// Loop and create objects
if (!empty($users_data)) {
    foreach ($users_data as $data) {
        $user_objects[] = new User($data['name'], $data['email']);
    }
}

// Call function
processUsers($user_objects);

// Alternative syntax for control structures
$count = 0;
while ($count < 3):
    echo "Count is $count\n";
    $count++;
endwhile;

echo "Script completed.\n";

?>
