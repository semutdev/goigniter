<?php
defined('BASEPATH') OR exit('No direct script access allowed');

/*
| -------------------------------------------------------------------------
| URI ROUTING
| -------------------------------------------------------------------------
*/
$route['default_controller'] = 'users';
$route['404_override'] = '';
$route['translate_uri_dashes'] = FALSE;

// Custom application routes
$route['users'] = 'users/index';
$route['users/create'] = 'users/create';
$route['users/(:num)'] = 'users/detail/$1';
$route['api/users']['get'] = 'users/api_index';
