# frozen_string_literal: true

module Db
  class CharacterImagesController < Db::ApplicationController
    permits :attachment, :asin, :copyright

    before_action :authenticate_user!
    before_action :load_character, only: %i(show create update destroy)
    before_action :load_image, only: %i(update destroy)

    def show
      @image = @character.character_image.presence || @character.build_character_image
    end

    def create(character_image)
      @image = @character.build_character_image(character_image)
      authorize @image, :create?
      @image.user = current_user

      if @image.save
        flash[:notice] = t "messages.character_images.saved"
        redirect_to db_character_image_path(@character)
      else
        render :show
      end
    end

    def update(character_image)
      authorize @image, :update?

      @image.attributes = character_image
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
  end
end
