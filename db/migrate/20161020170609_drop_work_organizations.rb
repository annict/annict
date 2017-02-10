# frozen_string_literal: true

class DropWorkOrganizations < ActiveRecord::Migration[5.0]
  def change
    drop_table :draft_work_organizations
    drop_table :work_organizations
  end
end
