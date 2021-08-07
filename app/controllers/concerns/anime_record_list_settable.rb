# frozen_string_literal: true

module AnimeRecordListSettable
  extend ActiveSupport::Concern

  def set_anime_record_list(anime)
    records = anime
      .records_only_anime
      .only_kept
      .eager_load(:anime, :anime_record, :episode_record, user: %i[gumroad_subscriber profile setting])
      .merge(AnimeRecord.only_kept.order_by_rating(:desc).order(created_at: :desc))
    @my_records = @following_records = []

    if user_signed_in?
      @my_records = records.merge(current_user.records.only_kept)
      @following_records = records.merge(current_user.followings)
      @all_records = records
        .where.not(user: [current_user, *current_user.followings])
        .merge(AnimeRecord.with_body)
        .page(params[:page])
        .per(100)
        .without_count
    else
      @all_records = records
        .page(params[:page])
        .per(100)
        .without_count
    end
  end
end
