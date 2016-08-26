# frozen_string_literal: true

class RecordsListService
  def initialize(episode, current_user, record_user)
    @episode = episode
    @current_user = current_user
    @record_user = record_user
  end

  def record_user_ids
    record_users = User.joins(:checkins).
      where("checkins.episode_id": @episode.id).
      where("checkins.user_id": base_records.pluck(:user_id).uniq).
      order("checkins.id DESC")
    record_users.pluck(:id).uniq
  end

  def user_records
    if @record_user.present?
      base_records.where(user: @record_user)
    else
      base_records.none
    end
  end

  def current_user_records
    if @current_user.present?
      base_records.where(user: @current_user).order(created_at: :desc)
    else
      base_records.none
    end
  end

  def records
    results = base_records.where.not(id: record_ids)

    if @current_user.present?
      mute_user_ids = @current_user.mute_users.pluck(:muted_user_id)
      results = results.where.not(user_id: mute_user_ids)
    end

    results.order(created_at: :desc)
  end

  private

  def base_records
    @episode.checkins.includes(user: :profile)
  end

  def record_ids
    user_records.pluck(:id) | current_user_records.pluck(:id)
  end
end
