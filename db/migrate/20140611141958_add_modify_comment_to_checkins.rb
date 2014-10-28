class AddModifyCommentToCheckins < ActiveRecord::Migration
  def change
    add_column :checkins, :modify_comment, :boolean, default: false, null: false, after: :spoil
  end
end
