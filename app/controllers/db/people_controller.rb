class Db::PeopleController < Db::ApplicationController
  before_action :authenticate_user!, only: [:new, :create, :edit, :update, :destroy]

  def index(page: nil)
    @people = Person.order(id: :desc).page(page)
  end
end
