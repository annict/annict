# frozen_string_literal: true

namespace :tmp do
  task update_for_changing_schema: :environment do
    DbActivity.where(action: "pvs.create").update_all(trackable_type: "Trailer", action: "trailers.create")
    DbActivity.where(action: "pvs.update").update_all(trackable_type: "Trailer", action: "trailers.update")

    DbActivity.where(action: "programs.create").update_all(trackable_type: "Slot", action: "slots.create")
    DbActivity.where(action: "programs.update").update_all(trackable_type: "Slot", action: "slots.update")

    DbActivity.where(action: "program_details.create").update_all(trackable_type: "Program", action: "programs.create")
    DbActivity.where(action: "program_details.update").update_all(trackable_type: "Program", action: "programs.update")
  end
end
