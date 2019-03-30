# frozen_string_literal: true

module Db
  class CharactersController < Db::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update)
    before_action :load_character, only: %i(edit update activities hide destroy)

    def index
      @characters = Character.order(id: :desc).page(params[:page])
    end

    def new
      @form = DB::CharacterRowsForm.new
      authorize @form, :new?
    end

    def create
      @form = DB::CharacterRowsForm.new(character_rows_form_params)
      @form.user = current_user
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_characters_path, notice: t("resources.character.created")
    end

    def edit
      authorize @character, :edit?
    end

    def update
      authorize @character, :update?

      @character.attributes = character_params
      @character.user = current_user

      return render(:edit) unless @character.valid?
      @character.save_and_create_activity!

      message = t("resources.character.updated")
      redirect_to db_characters_path, notice: message
    end

    def hide
      authorize @character, :hide?

      @character.hide!

      flash[:notice] = t("messages._common.updated")
      redirect_back fallback_location: db_characters_path
    end

    def destroy
      authorize @character, :destroy?

      @character.destroy

      flash[:notice] = t("messages._common.deleted")
      redirect_back fallback_location: db_characters_path
    end

    def activities
      @activities = @character.db_activities.order(id: :desc)
      @comment = @character.db_comments.new
    end

    private

    def load_character
      @character = Character.find(params[:id])
    end

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
