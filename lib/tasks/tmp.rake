# frozen_string_literal: true

namespace :tmp do
  task rename_checkin_to_record: :environment do
    Activity.where(trackable_type: "Checkin").update_all(trackable_type: "Record")
    Like.where(recipient_type: "Checkin").update_all(recipient_type: "Record")
  end
end
