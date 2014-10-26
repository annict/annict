create_table "users", force: true do |t|
  t.string   "username",             limit: 510,                 null: false
  t.string   "email",                limit: 510,                 null: false
  t.integer  "role",                                             null: false
  t.string   "encrypted_password",   limit: 510, default: "",    null: false
  t.datetime "remember_created_at"
  t.integer  "sign_in_count",                    default: 0,     null: false
  t.datetime "current_sign_in_at"
  t.datetime "last_sign_in_at"
  t.string   "current_sign_in_ip",   limit: 510
  t.string   "last_sign_in_ip",      limit: 510
  t.string   "confirmation_token",   limit: 510
  t.datetime "confirmed_at"
  t.datetime "confirmation_sent_at"
  t.string   "unconfirmed_email",    limit: 510
  t.integer  "checkins_count",                   default: 0,     null: false
  t.integer  "notifications_count",              default: 0,     null: false
  t.datetime "created_at"
  t.datetime "updated_at"
  t.boolean  "share_checkin",                    default: false
end
