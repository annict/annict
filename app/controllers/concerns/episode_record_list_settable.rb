# frozen_string_literal: true

module EpisodeRecordListSettable
  extend ActiveSupport::Concern

  def set_episode_record_list(episode)
    records = episode
      .records
      .preload(:work, :work_record, episode_record: :episode)
      .eager_load(user: %i[gumroad_subscriber profile setting])
      .only_kept
    @my_records = @following_records = @all_records = Record.none

    if user_signed_in?
      @my_records = current_user
        .records
        .preload(:work, :episode)
        .only_kept
        .where(episode: episode)
        .order(created_at: :desc)
      @following_records = records
        .merge(current_user.followings)
        .order(created_at: :desc)
      @all_records = records
        .where.not(user: [current_user, *current_user.followings])
    else
      @all_records = records
    end

    @all_records = @all_records
      .with_body
      .order_by_rating(:desc)
      .page(params[:page])
      .per(20)
  end
end
