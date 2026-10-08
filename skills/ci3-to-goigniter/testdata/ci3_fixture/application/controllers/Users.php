<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class Users extends CI_Controller {

    public function __construct() {
        parent::__construct();
        $this->load->model('User_model');
        $this->load->helper('url');
        $this->load->helper('form');
    }

    public function index() {
        $data['title'] = 'Users List';
        $data['users'] = $this->User_model->get_all();
        $this->load->view('users/index', $data);
    }

    public function detail($id) {
        $data['title'] = 'User Detail';
        $data['user'] = $this->User_model->get_by_id($id);
        $this->load->view('users/detail', $data);
    }

    public function create() {
        $data['title'] = 'Create User';
        if ($this->input->method() === 'post') {
            $userData = array(
                'name' => $this->input->post('name'),
                'email' => $this->input->post('email'),
                'is_active' => 1,
            );
            $this->User_model->insert_user($userData);
            redirect('users');
        }
        $this->load->view('users/create', $data);
    }
}
