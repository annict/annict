module Db
  class CastParticipationsController < Db::ApplicationController
    permits :work_id, :name, :character_name

    before_action :authenticate_user!
    before_action :load_person, only: [:index, :new, :create]

    def index
      @cast_participations = @person.cast_participations.order(id: :desc)
    end

    def new
      @cast_participation = @person.cast_participations.new
      authorize @cast_participation, :new?
    end

    def create(cast_participation)
      @cast_participation = @person.cast_participations.new(cast_participation)
      authorize @cast_participation, :create?

      if @cast_participation.valid?
        key = "cast_participations.create"
        @cast_participation.save_and_create_db_activity(current_user, key)
        redirect_to db_person_cast_participations_path(@person), notice: "登録しました"
      else
        render :new
      end
    end

    private

    def load_person
      @person = Person.find(params[:person_id])
    end
  end
end
