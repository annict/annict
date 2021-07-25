# frozen_string_literal: true

module Api::Internal
  class CommentedAnimeRecordsController < ApplicationV6Controller
    before_action :authenticate_user!

    def create
      @anime = Anime.only_kept.find(params[:anime_id])
      @form = Forms::AnimeRecordForm.new(anime: @anime, **anime_record_form_params)

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      Creators::AnimeRecordCreator.new(user: current_user, form: @form).call

      render(json: {}, status: 201)
    end

    private

    def anime_record_form_params
      params.require(:forms_anime_record_form).permit(:comment, :rating_overall, :share_to_twitter)
    end
  end
end
