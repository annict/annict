# frozen_string_literal: true

class RecordDecorator < ApplicationDecorator
  def records_show_title
    if episode.present?
      I18n.t("head.title.records.show_episode",
        profile_name: profile_name, username: username, work_title: work_title, episode_title_with_number: episode_title_with_number)
    else
      I18n.t("head.title.records.show_work", profile_name: profile_name, username: username, work_title: work_title)
    end
  end

  def records_show_description
    if episode.present?
      I18n.t("head.meta.description.records.show_episode",
        profile_name: profile_name, username: username, work_title: work_title, episode_title_with_number: episode_title_with_number)
    else
      I18n.t("head.meta.description.records.show_work",
        profile_name: profile_name, username: username, work_title: work_title)
    end
  end

  private

  def profile_name
    user.profile.name
  end

  def username
    user.username
  end

  def work_title
    work.decorate.local_title
  end

  def episode_title_with_number
    episode.decorate.title_with_number
  end
end
