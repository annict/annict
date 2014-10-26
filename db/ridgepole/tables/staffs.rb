create_table "staffs", force: true do |t|
  t.string   "email",              limit: 510, default: "", null: false
  t.string   "encrypted_password", limit: 510, default: "", null: false
  t.integer  "sign_in_count",                  default: 0,  null: false
  t.datetime "current_sign_in_at"
  t.datetime "last_sign_in_at"
  t.string   "current_sign_in_ip", limit: 510
  t.string   "last_sign_in_ip",    limit: 510
  t.datetime "created_at"
  t.datetime "updated_at"
end
