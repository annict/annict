module Db
  class CastsController < Db::ApplicationController
    permits :work_id, :name, :part

    before_action :authenticate_user!
    before_action :load_person, only: [:index, :new, :create, :edit, :update]

    def index
      @casts = @person.casts.order(id: :desc)
    end

    def new
      @cast = @person.casts.new
      authorize @cast, :new?
    end

    def create(cast)
      @cast = @person.casts.new(cast)
      authorize @cast, :create?

      if @cast.valid?
        key = "casts.create"
        @cast.save_and_create_db_activity(current_user, key)
        redirect_to db_person_casts_path(@person), notice: "登録しました"
      else
        render :new
      end
    end

    def edit(id)
      @cast = @person.casts.find(id)
      authorize @cast, :edit?
    end

    def update(id, cast)
      @cast = @person.casts.find(id)
      authorize @cast, :update?
      @cast.attributes = cast

      if @cast.valid?
        key = "casts.update"
        @cast.save_and_create_db_activity(current_user, key)
        redirect_to db_person_casts_path(@person), notice: "更新しました"
      else
        render :edit
      end
    end

    private

    def load_person
      @person = Person.find(params[:person_id])
    end
  end
end
