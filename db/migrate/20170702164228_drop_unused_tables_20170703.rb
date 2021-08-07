# frozen_string_literal: true

class DropUnusedTables20170703 < ActiveRecord::Migration[5.1]
  def change
    %i[
      cover_images
      draft_casts
      draft_episodes
      draft_multiple_episodes
      draft_organizations
      draft_people
      draft_programs
      draft_staffs
      draft_works
      edit_request_comments
      edit_request_participants
      edit_requests
    ].each do |table_name|
      drop_table table_name, force: true
    end
  end
end
