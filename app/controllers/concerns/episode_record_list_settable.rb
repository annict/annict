# frozen_string_literal: true

module EpisodeRecordListSettable
  extend ActiveSupport::Concern

  def set_episode_record_list(episode)
    records = episode
      .records
      .preload(:work, episode_record: :episode)
      .eager_load(user: %i[gumroad_subscriber profile setting])
      .only_kept
      .merge(EpisodeRecord.order_by_rating_state(:desc))
      .order(watched_at: :desc)
    @my_records = @following_records = Record.none

    if user_signed_in?
      records = records
        .where.not(user_id: current_user.mute_users.pluck(:muted_user_id))
      @my_records = current_user
        .records
        .eager_load(episode_record: :episode)
        .preload(:work)
        .only_kept
        .where(episode_records: {episodes: episode})
        .order(watched_at: :desc)
      @following_records = records
        .merge(current_user.followings)
      @all_records = records
        .where.not(user: [current_user, *current_user.followings])
    else
      @all_records = records
    end

    @all_records = @all_records
      .merge(EpisodeRecord.with_body)
      .page(params[:page])
      .per(20)
  end
end
