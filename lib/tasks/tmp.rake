# frozen_string_literal: true

namespace :tmp do
  task change_db_activities_action: :environment do
    DbActivity.where(action: "programs.create").update_all(action: "slots.create")
    DbActivity.where(action: "programs.update").update_all(action: "slots.update")
  end
end
