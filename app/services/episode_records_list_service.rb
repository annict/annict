# frozen_string_literal: true

class EpisodeRecordsListService
  PAGE_PER = 30

  def initialize(user, episode, params)
    @user = user
    @episode = episode
    @params = params
  end

  def all_comment_episode_records
    results = all_episode_records
    results = results.with_comment
    results = localable_episode_records(results)
    results = results.page(@params[:page])
    results = sort(results)
    results.per(PAGE_PER)
  end

  def friend_comment_episode_records
    return EpisodeRecord.none if @user.blank?

    results = all_episode_records
    results = results.with_comment
    results = results.joins(:user).merge(@user.followings.published)
    results = localable_episode_records(results)
    results = results.page(@params[:page])
    results = sort(results)
    results.per(PAGE_PER)
  end

  def my_episode_records
    return EpisodeRecord.none if @user.blank?

    results = all_episode_records
    results = results.where(user: @user)
    results = localable_episode_records(results)
    results = results.page(@params[:page])
    results = sort(results)
    results.per(PAGE_PER)
  end

  def selected_comment_episode_records
    return all_comment_episode_records if @user.blank?

    case @user.setting.display_option_record_list
    when "all_comments" then all_comment_episode_records
    when "friend_comments" then friend_comment_episode_records
    when "my_episode_records" then my_episode_records
    end
  end

  def all_episode_records
    UserEpisodeRecordsQuery.new.call(
      episode_records:@episode.episode_records,
      user: @user
    )
  end

  private

  def sort(results)
    return results if @user.blank?

    case @user.setting.records_sort_type
    when "likes_count_desc"
      results.order(likes_count: :desc).order(created_at: :desc)
    when "rating_state_desc"
      results.rating_state_order(:desc).order(created_at: :desc)
    when "rating_state_asc"
      results.rating_state_order(:asc).order(created_at: :desc)
    when "created_at_desc"
      results.order(created_at: :desc)
    when "created_at_asc"
      results.order(created_at: :asc)
    end
  end

  def localable_episode_records(episode_records)
    if @user.present?
      episode_records.where(user: @user).or(episode_records.with_locale(*@user.allowed_locales))
    elsif @user.blank? && @params[:locale_en]
      episode_records.with_locale(:en)
    elsif @user.blank? && @params[:locale_ja]
      episode_records.with_locale(:ja)
    else
      episode_records
    end
  end
end
