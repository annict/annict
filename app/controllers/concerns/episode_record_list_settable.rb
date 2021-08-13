# frozen_string_literal: true

module EpisodeRecordListSettable
  extend ActiveSupport::Concern

  def set_episode_record_list(episode)
    records = episode.records
      .only_kept
      .preload(:anime, :anime_record, episode_record: :episode)
      .eager_load(user: %i[gumroad_subscriber profile setting])
      .merge(EpisodeRecord.with_body.order_by_rating_state(:desc).order(created_at: :desc))
    @my_records = @following_records = Record.none

    if user_signed_in?
      @my_records = current_user
        .records
        .only_kept
        .eager_load(:episode_record)
        .where(episode_records: {episode_id: episode.id})
        .order(created_at: :desc)
      @following_records = records.merge(current_user.followings)
      @all_records = records.where.not(user: [current_user, *current_user.followings]).page(params[:page]).per(20)
    else
      @all_records = records.page(params[:page]).per(20)
    end
  end
end
