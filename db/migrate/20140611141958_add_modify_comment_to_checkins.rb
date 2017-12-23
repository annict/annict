class AddModifyCommentToCheckins < ActiveRecord::Migration[4.2]
  def change
    add_column :checkins, :modify_comment, :boolean, default: false, null: false, after: :spoil
  end
end
