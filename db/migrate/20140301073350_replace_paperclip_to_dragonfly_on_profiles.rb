class ReplacePaperclipToDragonflyOnProfiles < ActiveRecord::Migration
  def change
    add_column :profiles, :avatar_uid, :string, after: :description
  end
end
