# frozen_string_literal: true

class UserWorksQuery
  def initialize(user)
    @user = user
  end

  def watching
    Anime.joins(:library_entries).merge(@user.library_entries.watching)
  end

  def wanna_watch_and_watching
    Anime.joins(:library_entries).merge(@user.library_entries.wanna_watch_and_watching)
  end

  def desiring_to_watch
    Anime.joins(:library_entries).merge(@user.library_entries.desiring_to_watch)
  end

  def on_hold
    Anime.joins(:library_entries).merge(@user.library_entries.on_hold)
  end

  def unknown
    Anime.where.not(id: @user.library_entries.pluck(:work_id))
  end

  def on(status_kind)
    Anime.joins(:library_entries).merge(@user.library_entries.with_status(status_kind))
  end

  def all
    Anime.joins(:library_entries).merge(@user.library_entries)
  end
end
