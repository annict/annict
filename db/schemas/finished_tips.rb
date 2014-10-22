create_table "finished_tips", force: true do |t|
  t.integer  "user_id",    null: false
  t.integer  "tip_id",     null: false
  t.datetime "created_at", null: false
  t.datetime "updated_at", null: false
end

add_foreign_key "finished_tips", "tips", name: "finished_tips_tip_id_fk", dependent: :delete
add_foreign_key "finished_tips", "users", name: "finished_tips_user_id_fk", dependent: :delete
