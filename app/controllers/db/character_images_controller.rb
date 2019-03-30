# frozen_string_literal: true

module Db
  class CharacterImagesController < Db::ApplicationController
    before_action :authenticate_user!
    before_action :load_character, only: %i(show create update destroy)
    before_action :load_image, only: %i(update destroy)

    def show
      @image = @character.character_image.presence || @character.build_character_image
    end

    def create
      @image = @character.build_character_image(character_image_params)
      authorize @image, :create?
      @image.user = current_user

      if @image.save
        flash[:notice] = t "messages.character_images.saved"
        redirect_to db_character_image_path(@character)
      else
        render :show
      end
    end

    def update
      authorize @image, :update?

      @image.attributes = character_image_params
      @image.user = current_user

      if @image.save
        flash[:notice] = t "messages.character_images.saved"
        redirect_to db_character_image_path(@character)
      else
        render :show
      end
    end

    def destroy
      authorize @item, :destroy?
      @item.destroy
      redirect_to db_character_image_path(@character), notice: t("messages.character_images.deleted")
    end

    private

    def load_image
      @image = @character.character_image
      raise ActiveRecord::RecordNotFound if @image.blank?
    end

    def character_image_params
      params.require(:character_image).permit(:attachment, :asin, :copyright)
    end
  end
end
