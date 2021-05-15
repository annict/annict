# frozen_string_literal: true

module V6::EpisodeRecordListSettable
  extend ActiveSupport::Concern

  def set_episode_record_list(episode)
    records = episode.records.only_kept.eager_load(:episode_record, user: %i[gumroad_subscriber profile setting])
      .merge(EpisodeRecord.with_body.order_by_rating_state(:desc).order(created_at: :desc))
    @my_records = @following_records = []

    if user_signed_in?
      is_tracked = current_user.episode_records.only_kept.where(episode_id: episode.id).exists?
      likes = current_user.likes.select(:recipient_id, :recipient_type)

      @my_records = current_user
        .records
        .only_kept
        .eager_load(:episode_record)
        .where(episode_records: {episode_id: episode.id})
        .order(created_at: :desc)
      @my_records.each do |record|
        record.is_spoiler = false
        record.is_liked = record.liked?(likes)
      end

      @following_records = records.merge(current_user.followings)
      @following_records.each do |record|
        record.is_spoiler = !is_tracked
        record.is_liked = record.liked?(likes)
      end

      @all_records = records.where.not(user: [current_user, *current_user.followings]).page(params[:page]).per(3)
      @all_records.each do |record|
        record.is_spoiler = !is_tracked
        record.is_liked = record.liked?(likes)
      end
    else
      @all_records = records.page(params[:page]).per(30)
    end
  end
end
