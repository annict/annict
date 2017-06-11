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
    results = results.page(@params[:page])
    results = sort(results)
    results
  end

  def friend_comment_records
    return Checkin.none if @user.blank?

    results = all_records
    results = results.with_comment
    results = results.joins(:user).merge(@user.followings)
    results = results.page(@params[:page])
    results = sort(results)
    results
  end

  def my_comment_records
    return Checkin.none if @user.blank?

    results = all_records
    results = results.with_comment
    results = results.where(user: @user)
    results = results.page(@params[:page])
    results = sort(results)
    results
  end

  def selected_comment_records
    return all_comment_records if @user.blank?

    case @user.setting.display_option_record_list
    when "all_comments" then all_comment_records
    when "friend_comments" then friend_comment_records
    when "my_comments" then my_comment_records
    end
  end

  def all_records
    @episode.records.includes(user: :profile)
  end

  private

  def sort(results)
    return results if @user.blank?

    case @user.setting.records_sort_type
    when "likes_count_desc"
      results.order(likes_count: :desc)
    when "rating_state_desc"
      results.rating_state_order(:desc)
    when "rating_state_asc"
      results.rating_state_order(:asc)
    when "created_at_desc"
      results.order(created_at: :desc)
    when "created_at_asc"
      results.order(created_at: :asc)
    end
  end
end
