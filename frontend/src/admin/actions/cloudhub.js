import _ from 'lodash'
import uuid from 'uuid'

import {
  getUsers as getUsersAJAX,
  getOrganizations as getOrganizationsAJAX,
  createUser as createUserAJAX,
  updateUser as updateUserAJAX,
  deleteUser as deleteUserAJAX,
  createOrganization as createOrganizationAJAX,
  updateOrganization as updateOrganizationAJAX,
  deleteOrganization as deleteOrganizationAJAX,
  getMappings as getMappingsAJAX,
  createMapping as createMappingAJAX,
  updateMapping as updateMappingAJAX,
  deleteMapping as deleteMappingAJAX,
  createCloudServiceProvider as createCloudServiceProviderAJAX,
  updateCloudServiceProvider as updateCloudServiceProviderAJAX,
} from 'src/admin/apis/cloudhub'

import {notify} from 'src/shared/actions/notifications'
import {errorThrown} from 'src/shared/actions/errors'
import {
  notifyMappingDeleted,
  notifyCloudHubOrgDeleted,
  notifyCloudHubUserUpdated,
  notifyCloudHubUserDeleted,
  notifyCloudHubBasicUserAdd,
  notifyCloudHubOrgDeletionFailedWithRegisteredDevices,
} from 'src/shared/copy/notifications'

import {REVERT_STATE_DELAY} from 'src/shared/constants'

// action creators

// response contains `users` and `links`
export const loadUsers = ({users}) => ({
  type: 'CLOUDHUB_LOAD_USERS',
  payload: {
    users,
  },
})

export const loadOrganizations = ({organizations}) => ({
  type: 'CLOUDHUB_LOAD_ORGANIZATIONS',
  payload: {
    organizations,
  },
})

export const addUser = user => ({
  type: 'CLOUDHUB_ADD_USER',
  payload: {
    user,
  },
})

export const updateUser = (user, updatedUser) => ({
  type: 'CLOUDHUB_UPDATE_USER',
  payload: {
    user,
    updatedUser,
  },
})

export const syncUser = (staleUser, syncedUser) => ({
  type: 'CLOUDHUB_SYNC_USER',
  payload: {
    staleUser,
    syncedUser,
  },
})

export const removeUser = user => ({
  type: 'CLOUDHUB_REMOVE_USER',
  payload: {
    user,
  },
})

export const addOrganization = organization => ({
  type: 'CLOUDHUB_ADD_ORGANIZATION',
  payload: {
    organization,
  },
})

export const renameOrganization = (organization, newName) => ({
  type: 'CLOUDHUB_RENAME_ORGANIZATION',
  payload: {
    organization,
    newName,
  },
})

export const syncOrganization = (staleOrganization, syncedOrganization) => ({
  type: 'CLOUDHUB_SYNC_ORGANIZATION',
  payload: {
    staleOrganization,
    syncedOrganization,
  },
})

export const removeOrganization = organization => ({
  type: 'CLOUDHUB_REMOVE_ORGANIZATION',
  payload: {
    organization,
  },
})

export const loadMappings = ({mappings}) => ({
  type: 'CLOUDHUB_LOAD_MAPPINGS',
  payload: {
    mappings,
  },
})

export const updateMapping = (staleMapping, updatedMapping) => ({
  type: 'CLOUDHUB_UPDATE_MAPPING',
  payload: {
    staleMapping,
    updatedMapping,
  },
})

export const addMapping = mapping => ({
  type: 'CLOUDHUB_ADD_MAPPING',
  payload: {
    mapping,
  },
})

export const removeMapping = mapping => ({
  type: 'CLOUDHUB_REMOVE_MAPPING',
  payload: {
    mapping,
  },
})

export const createCloudServiceProviderAction = () => ({
  type: 'CREATE_CLOUD_SERVICE_PROVIDER',
})

export const updateCloudServiceProviderAction = () => ({
  type: 'UPDATE_CLOUD_SERVICE_PROVIDER',
})

// async actions (thunks)
export const loadUsersAsync = url => async dispatch => {
  try {
    const {data} = await getUsersAJAX(url)
    dispatch(loadUsers(data))
  } catch (error) {
    dispatch(errorThrown(error))
  }
}

export const loadOrganizationsAsync = url => async dispatch => {
  try {
    const {data} = await getOrganizationsAJAX(url)
    dispatch(loadOrganizations(data))
  } catch (error) {
    dispatch(errorThrown(error))
  }
}

export const loadMappingsAsync = () => async dispatch => {
  try {
    const {data} = await getMappingsAJAX()
    dispatch(loadMappings(data))
  } catch (error) {
    dispatch(errorThrown(error))
  }
}

export const createMappingAsync = (url, mapping) => async dispatch => {
  const mappingWithTempId = {...mapping, _tempID: uuid.v4()}
  dispatch(addMapping(mappingWithTempId))
  try {
    const {data} = await createMappingAJAX(url, mapping)
    dispatch(updateMapping(mappingWithTempId, data))
  } catch (error) {
    const message = `${_.upperFirst(_.toLower(error.data.message))}: Scheme: ${
      mapping.scheme
    } Provider: ${mapping.provider}`
    dispatch(errorThrown(error, message))
    setTimeout(
      () => dispatch(removeMapping(mappingWithTempId)),
      REVERT_STATE_DELAY
    )
  }
}

export const deleteMappingAsync = mapping => async dispatch => {
  dispatch(removeMapping(mapping))
  try {
    await deleteMappingAJAX(mapping)
    dispatch(notify(notifyMappingDeleted(mapping.id, mapping.scheme)))
  } catch (error) {
    dispatch(errorThrown(error))
    dispatch(addMapping(mapping))
  }
}

export const updateMappingAsync = (
  staleMapping,
  updatedMapping
) => async dispatch => {
  dispatch(updateMapping(staleMapping, updatedMapping))
  try {
    await updateMappingAJAX(updatedMapping)
  } catch (error) {
    dispatch(errorThrown(error))
    dispatch(updateMapping(updatedMapping, staleMapping))
  }
}

export const createUserAsync = (url, user) => async dispatch => {
  // temp uuid is added to be able to disambiguate a created user that has the
  // same scheme, provider, and name as an existing user
  const userWithTempID = {...user, _tempID: uuid.v4()}
  dispatch(addUser(userWithTempID))
  try {
    const {data} = await createUserAJAX(url, user)
    dispatch(syncUser(userWithTempID, data))
    dispatch(notify(notifyCloudHubBasicUserAdd(data.name, data.password)))
  } catch (error) {
    const message = `${_.upperFirst(_.toLower(error.data.message))}: ${
      user.scheme
    }::${user.provider}::${user.name}`
    dispatch(errorThrown(error, message))
    // undo optimistic update
    setTimeout(() => dispatch(removeUser(userWithTempID)), REVERT_STATE_DELAY)
  }
}

export const updateUserAsync = (
  user,
  updatedUser,
  successMessage
) => async dispatch => {
  dispatch(updateUser(user, updatedUser))
  try {
    // currently the request will be rejected if name, provider, or scheme, or
    // no roles are sent with the request.
    // TODO: remove the null assignments below so that the user request can have
    // the original name, provider, and scheme once the change to allow this is
    // implemented server-side
    const {data} = await updateUserAJAX({
      ...updatedUser,
      name: null,
      provider: null,
      scheme: null,
    })
    dispatch(notify(notifyCloudHubUserUpdated(successMessage)))
    // it's not necessary to syncUser again but it's useful for good
    // measure and for the clarity of insight in the redux story
    dispatch(syncUser(user, data))
  } catch (error) {
    dispatch(errorThrown(error))
    dispatch(syncUser(user, user))
  }
}

export const deleteUserAsync = (
  user,
  {isAbsoluteDelete} = {}
) => async dispatch => {
  dispatch(removeUser(user))
  try {
    await deleteUserAJAX(user)
    dispatch(notify(notifyCloudHubUserDeleted(user.name, isAbsoluteDelete)))
  } catch (error) {
    dispatch(errorThrown(error))
    dispatch(addUser(user))
  }
}

export const createOrganizationAsync = (
  url,
  organization
) => async dispatch => {
  // temp uuid is added to be able to disambiguate a created organization with
  // the same name as an existing organization
  const organizationWithTempID = {...organization, _tempID: uuid.v4()}
  dispatch(addOrganization(organizationWithTempID))
  try {
    const {data} = await createOrganizationAJAX(url, organization)
    dispatch(syncOrganization(organization, data))
  } catch (error) {
    const message = `${_.upperFirst(_.toLower(error.data.message))}: ${
      organization.name
    }`
    dispatch(errorThrown(error, message))
    // undo optimistic update
    setTimeout(
      () => dispatch(removeOrganization(organizationWithTempID)),
      REVERT_STATE_DELAY
    )
  }
}

export const updateOrganizationAsync = (
  organization,
  updatedOrganization
) => async dispatch => {
  dispatch(renameOrganization(organization, updatedOrganization.name))
  try {
    const {data} = await updateOrganizationAJAX(updatedOrganization)
    // it's not necessary to syncOrganization again but it's useful for good
    // measure and for the clarity of insight in the redux story
    dispatch(syncOrganization(updatedOrganization, data))
  } catch (error) {
    dispatch(errorThrown(error))
    dispatch(syncOrganization(organization, organization)) // restore if fail
  }
}

export const deleteOrganizationAsync = organization => async dispatch => {
  dispatch(removeOrganization(organization))
  try {
    await deleteOrganizationAJAX(organization)
    dispatch(notify(notifyCloudHubOrgDeleted(organization.name)))
  } catch (error) {
    if (error?.status === 409) {
      dispatch(
        notify(
          notifyCloudHubOrgDeletionFailedWithRegisteredDevices(
            organization?.name || ''
          )
        )
      )
      dispatch(addOrganization(organization))
      return
    }
    dispatch(errorThrown(error))
    dispatch(addOrganization(organization))
  }
}
