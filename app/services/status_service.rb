# frozen_string_literal: true

class StatusService
  attr_writer :app, :via, :ga_client, :page_category

  def initialize(user, work)
    @user = user
    @work = work
  end

  def change!(kind)
    @kind = kind

    if Status.kind.values.include?(@kind)
      @status = @user.statuses.new(work: @work, kind: @kind, oauth_application: @app)

      if @status.save!
        UserWatchedWorksCountJob.perform_later(@user)
        @status.share_to_sns
        create_ga_event
      end
    elsif @kind == "no_select"
      library_entry = @user.library_entries.find_by(work: @work)
      if library_entry
        library_entry.update_attribute(:status_id, nil)
        UserWatchedWorksCountJob.perform_later(@user)
      end
    end
  end

  private

  def create_ga_event
    return if @ga_client.blank?

    @ga_client.events.create(:statuses, :create, ds: @via)
  end
end
