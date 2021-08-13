# frozen_string_literal: true

module Updaters
  class StatusUpdater
    def initialize(user:, form:)
      @user = user
      @form = form
      @work = @form.work
    end

    def call
      library_entry = @user.library_entries.find_by(work: @work)
      prev_status_kind = library_entry&.status&.kind&.to_sym.presence || :no_select
      new_status_kind = Status.kind_v3_to_v2(@form.kind)

      ActiveRecord::Base.transaction do
        if @form.no_status?
          library_entry&.update!(status_id: nil)
        else
          status = @user.statuses.new(work: @work, kind: new_status_kind)

          status.save!
          status.save_library_entry!
          status.share_to_sns

          activity_group = @user.create_or_last_activity_group!(status)
          @user.activities.create!(itemable: status, activity_group: activity_group)
        end

        AfterStatusUpdateJob.perform_later(@user.id, @work.id, prev_status_kind, new_status_kind)
      end
    end
  end
end
