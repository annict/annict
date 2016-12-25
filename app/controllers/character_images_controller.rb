# frozen_string_literal: true

class CharacterImagesController < ApplicationController
  permits :attachment, :source_url

  before_action :authenticate_user!, only: %i(new create destroy)
  before_action :load_character, only: %i(index new create destroy)
  before_action :load_i18n, only: %i(index)

  def index
    @images = @character.character_images.published.order(updated_at: :desc)
  end

  def new
    @image = @character.character_images.new
  end

  def create(character_image)
    @image = @character.character_images.new(character_image)
    @image.user = current_user

    if @image.save
      flash[:notice] = t "messages.character_images.uploaded"
      redirect_to character_images_path(@character)
    else
      render :new
    end
  end

  def destroy(id)
    @image = @character.character_images.find(id)

    authorize @image, :destroy?
    if @character.character_image == @image
      return redirect_to :back, alert: t("messages.character_images.can_not_delete")
    end

    @image.destroy
    flash[:notice] = t "messages.character_images.deleted"
    redirect_to character_images_path(@character)
  end

  private

  def load_i18n
    keys = {
      "messages.components.thumbs_buttons.require_sign_in": nil,
      "messages.components.thumbs_buttons.can_not_vote_to_owned_image": nil
    }
    load_i18n_into_gon keys
  end
end
