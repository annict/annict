# frozen_string_literal: true

module My
  class EpisodeRecordsController < V4::ApplicationController
    layout false

    before_action :authenticate_user!, only: %i(create)

    def index
      episode = Episode.only_kept.find(params[:episode_id])
      records = episode.records.only_kept.eager_load(:episode_record, user: %i(gumroad_subscriber profile setting)).
        merge(EpisodeRecord.with_body.order_by_rating_state(:desc).order(created_at: :desc))
      @my_records = @following_records = []

      if user_signed_in?
        is_tracked = current_user.episode_records.only_kept.where(episode_id: episode.id).exists?
        likes = current_user.likes.select(:recipient_id, :recipient_type)

        @my_records = current_user.
          records.
          only_kept.
          eager_load(:episode_record).
          where(episode_records: { episode_id: episode.id }).
          order(created_at: :desc)
        @my_records.each do |record|
          record.is_spoiler = false
          record.is_liked = record.liked?(likes)
        end

        @following_records = records.merge(current_user.followings)
        @following_records.each do |record|
          record.is_spoiler = !is_tracked
          record.is_liked = record.liked?(likes)
        end

        @all_records = records.where.not(user: current_user).page(params[:page]).per(30)
        @all_records.each do |record|
          record.is_spoiler = !is_tracked
          record.is_liked = record.liked?(likes)
        end
      else
        @all_records = records.page(params[:page]).per(30)
      end
    end

    def create
      @form = EpisodeRecordForm.new(episode_record_form_params)
      @form.episode = Episode.only_kept.find(params[:episode_id])

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      EpisodeRecordCreator2.new(user: current_user, form: @form).call

      head 201
    end

    private

    def episode_record_form_params
      params.required(:episode_record_form).permit(:comment, :rating, :share_to_twitter)
    end
  end
end
