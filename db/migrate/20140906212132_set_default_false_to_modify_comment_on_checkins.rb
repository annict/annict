class SetDefaultFalseToModifyCommentOnCheckins < ActiveRecord::Migration
  def change
    change_column_default :checkins, :modify_comment, :false
  end
end
