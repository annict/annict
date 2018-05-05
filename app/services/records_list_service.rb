# frozen_string_literal: true

class RecordsListService
  def initialize(user, episode, params)
    @user = user
    @episode = episode
    @params = params
  end

  def all_comment_records
    results = all_records
    results = results.with_comment
    results = localable_records(results)
    results = results.page(@params[:page])
    results = sort(results)
    results.per(20)
  end

  def friend_comment_records
    return Record.none if @user.blank?

    results = all_records
    results = results.with_comment
    results = results.joins(:user).merge(@user.followings)
    results = localable_records(results)
    results = results.page(@params[:page])
    results = sort(results)
    results.per(20)
  end

  def my_records
    return Record.none if @user.blank?

    results = all_records
    results = results.where(user: @user)
    results = localable_records(results)
    results = results.page(@params[:page])
    results = sort(results)
    results.per(20)
  end

  def selected_comment_records
    return all_comment_records if @user.blank?

    case @user.setting.display_option_record_list
    when "all_comments" then all_comment_records
    when "friend_comments" then friend_comment_records
    when "my_records" then my_records
    end
  end

  def all_records
    records = @episode.records.includes(user: :profile)
    records = records.where.not(user_id: @user.mute_users.pluck(:muted_user_id)) if @user.present?
    records
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

  def localable_records(records)
    if @user.present?
      records.where(user: @user).or(records.with_locale(*@user.allowed_locales))
    elsif @user.blank? && @params[:locale_en]
      records.with_locale(:en)
    elsif @user.blank? && @params[:locale_ja]
      records.with_locale(:ja)
    else
      records
    end
  end
end
