# typed: false
# frozen_string_literal: true

class Deprecated::UserWorksQuery
  def initialize(user)
    @user = user
  end

  def watching
    Work.joins(:library_entries).merge(@user.library_entries.watching)
  end

  def wanna_watch_and_watching
    Work.joins(:library_entries).merge(@user.library_entries.wanna_watch_and_watching)
  end

  def desiring_to_watch
    Work.joins(:library_entries).merge(@user.library_entries.desiring_to_watch)
  end

  def on_hold
    Work.joins(:library_entries).merge(@user.library_entries.on_hold)
  end

  def unknown
    Work.where.not(id: @user.library_entries.pluck(:work_id))
  end

  def on(status_kind)
    Work.joins(:library_entries).merge(@user.library_entries.with_status(status_kind))
  end

  def all
    Work.joins(:library_entries).merge(@user.library_entries)
  end
end
