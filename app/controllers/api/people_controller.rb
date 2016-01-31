module Api
  class PeopleController < Api::ApplicationController
    def index(q: nil)
      @people = if q.present?
        Person.where("name ILIKE ?", "%#{q}%").published
      else
        Person.none
      end
    end
  end
end
