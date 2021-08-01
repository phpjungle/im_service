<?php
const APP_ID = 7;
$data = file_get_contents("php://input");

$data = json_decode($data, true);

// {"data":{"token":"f64fdae9aa2e536e36becef55850b01d","cache_token":true}}
if (isset($data['uid'])) {
    $chat_id = $data['uid'];
    $token = md5(sprintf("%s_%s", APP_ID, $chat_id));
    $resp = ['data' => ['token' => $token, 'cache_token' => set_access_token($chat_id)]];

    echo json_encode($resp);
} else {
    $resp = ['data' => ['token' => '']];
    echo json_encode($resp);
}

function set_access_token($chat_id) {
    if ($chat_id) {
        $host = 'localhost';
        $redis = new PJRedis($host, 6379, null);
        $info = $redis->info();

        // var_dump($info);

        $redis->select(0);
        $key = sprintf("access_token_%s", md5(sprintf("%s_%s", APP_ID, $chat_id)));

        $stat = $redis->hash_sets($key, ['app_id' => APP_ID, 'user_id' => $chat_id]);

        return $stat;
    }
}