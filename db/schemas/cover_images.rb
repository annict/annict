create_table "cover_images", force: true do |t|
  t.integer  "work_id",                null: false
  t.string   "file_name",  limit: 510, null: false
  t.string   "location",   limit: 510, null: false
  t.datetime "created_at"
  t.datetime "updated_at"
end

add_index "cover_images", ["work_id"], name: "cover_images_work_id_idx", using: :btree

add_foreign_key "cover_images", "works", name: "cover_images_work_id_fk", dependent: :delete
