class SetDefaultFalseToModifyCommentOnCheckins < ActiveRecord::Migration[4.2]
  def change
    change_column_default :checkins, :modify_comment, :false
  end
end
