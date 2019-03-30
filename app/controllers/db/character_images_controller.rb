# frozen_string_literal: true

module Db
  class CharacterImagesController < Db::ApplicationController
    before_action :authenticate_user!

    def show
      @character = Character.find(params[:character_id])
      @image = @character.character_image.presence || @character.build_character_image
    end

    def create
      @character = Character.find(params[:character_id])
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
      @character = Character.find(params[:character_id])
      @image = CharacterImage.find_by!(character_id: @character.id)
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
      @character = Character.find(params[:character_id])
      @image = CharacterImage.find_by!(character_id: @character.id)
      authorize @item, :destroy?
      @item.destroy
      redirect_to db_character_image_path(@character), notice: t("messages.character_images.deleted")
    end

    private

    def character_image_params
      params.require(:character_image).permit(:attachment, :asin, :copyright)
    end
  end
end
