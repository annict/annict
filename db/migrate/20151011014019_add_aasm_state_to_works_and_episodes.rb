class AddAasmStateToWorksAndEpisodes < ActiveRecord::Migration
  def change
    add_column :works, :aasm_state, :string, null: false, default: "published"
    add_column :episodes, :aasm_state, :string, null: false, default: "published"

    add_index :works, :aasm_state
    add_index :episodes, :aasm_state
  end
end
