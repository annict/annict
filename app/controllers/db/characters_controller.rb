# frozen_string_literal: true

module Db
  class CharactersController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @characters = Character
        .without_deleted
        .preload(:series)
        .order(id: :desc)
        .page(params[:page])
        .per(100)
    end

    def new
      @form = Deprecated::Db::CharacterRowsForm.new
      authorize @form
    end

    def create
      @form = Deprecated::Db::CharacterRowsForm.new(character_rows_form_params)
      @form.user = current_user
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_character_list_path, notice: t("resources.character.created")
    end

    def edit
      @character = Character.without_deleted.find(params[:id])
      authorize @character
    end

    def update
      @character = Character.without_deleted.find(params[:id])
      authorize @character

      @character.attributes = character_params
      @character.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @character.valid?

      @character.save_and_create_activity!

      redirect_to db_edit_character_path(@character), notice: t("resources.character.updated")
    end

    def destroy
      @character = Character.without_deleted.find(params[:id])
      authorize @character

      @character.destroy_in_batches

      redirect_back(
        fallback_location: db_character_list_path,
        notice: t("messages._common.deleted")
      )
    end

    private

    def character_rows_form_params
      params.require(:db_character_rows_form).permit(:rows)
    end

    def character_params
      params.require(:character).permit(
        :name, :name_kana, :name_en, :series_id, :nickname, :nickname_en,
        :birthday, :birthday_en, :age, :age_en, :blood_type, :blood_type_en, :height,
        :height_en, :weight, :weight_en, :nationality, :nationality_en, :occupation,
        :occupation_en, :description, :description_en, :description_source,
        :description_source_en
      )
    end
  end
end
