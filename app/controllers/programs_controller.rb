class ProgramsController < ApplicationController
  before_action :authenticate_user!
end