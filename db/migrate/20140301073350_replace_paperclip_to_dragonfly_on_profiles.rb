class ReplacePaperclipToDragonflyOnProfiles < ActiveRecord::Migration[4.2]
  def change
    add_column :profiles, :avatar_uid, :string, after: :description
  end
end
