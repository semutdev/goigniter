<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class User_model extends CI_Model {

    protected $table = 'users';

    public function __construct() {
        parent::__construct();
        $this->load->database();
    }

    public function get_all() {
        $query = $this->db->get($this->table);
        return $query->result();
    }

    public function get_by_id($id) {
        $query = $this->db->get_where($this->table, array('id' => $id));
        return $query->row();
    }

    public function insert_user($data) {
        return $this->db->insert($this->table, $data);
    }

    public function update_user($id, $data) {
        $this->db->where('id', $id);
        return $this->db->update($this->table, $data);
    }

    public function delete_user($id) {
        $this->db->where('id', $id);
        return $this->db->delete($this->table);
    }
}
