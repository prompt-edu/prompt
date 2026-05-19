import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Save, X, Loader2, AlertCircle } from 'lucide-react'
import { DeleteDialog } from './DeleteDialog'
import {
  Button,
  Input,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Badge,
  Separator,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'

interface Entity {
  id: string
  name: string
}

interface EntitySettingsProps<T extends Entity> {
  items: T[]
  // mutation functions for create, update, delete operations
  createFn: (names: string[]) => Promise<any>
  updateFn: (id: string, newName: string) => Promise<any>
  deleteFn: (id: string) => Promise<any>
  // used to invalidate the cache for this entity type
  queryKey: any[]
  // display texts and icons
  title: string
  description: string
  icon: React.ReactNode
  emptyIcon: React.ReactNode
  emptyMessage: string
  emptySubtext: string
}

export const EntitySettings = <T extends Entity>({
  items,
  createFn,
  updateFn,
  deleteFn,
  queryKey,
  title,
  description,
  icon,
  emptyIcon,
  emptyMessage,
  emptySubtext,
}: EntitySettingsProps<T>) => {
  const queryClient = useQueryClient()

  const invalidateCache = () => {
    queryClient.invalidateQueries({ queryKey })
  }

  const createMutation = useMutation({
    mutationFn: (names: string[]) => createFn(names),
    onSuccess: () => {
      invalidateCache()
      setCreateError(null)
    },
    onError: () => {
      setCreateError(`Failed to create ${title.toLowerCase()}. Please try again.`)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, newName }: { id: string; newName: string }) => updateFn(id, newName),
    onSuccess: () => {
      invalidateCache()
      setUpdateError(null)
    },
    onError: () => {
      setUpdateError(`Failed to update ${title.toLowerCase()}. Please try again.`)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteFn(id),
    onSuccess: () => {
      invalidateCache()
      setDeleteError(null)
    },
    onError: () => {
      setDeleteError(`Failed to delete ${title.toLowerCase()}. Please try again.`)
    },
  })

  const [newName, setNewName] = useState('')
  const [editingItem, setEditingItem] = useState<{ id: string; name: string } | null>(null)
  const [createError, setCreateError] = useState<string | null>(null)
  const [updateError, setUpdateError] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  // sort items by name
  const sortedItems = [...items].sort((a, b) => a.name.localeCompare(b.name))

  const handleAddItem = () => {
    if (newName.trim()) {
      setCreateError(null)
      createMutation.mutate([newName.trim()])
      setNewName('')
    }
  }

  const handleUpdateItem = () => {
    if (editingItem && editingItem.name.trim()) {
      setUpdateError(null)
      updateMutation.mutate({ id: editingItem.id, newName: editingItem.name.trim() })
      setEditingItem(null)
    }
  }

  const handleDeleteItem = (id: string) => {
    setDeleteError(null)
    deleteMutation.mutate(id)
  }

  return (
    <Card className='w-full shadow-xs'>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div>
            <CardTitle className='text-2xl font-bold flex items-center gap-2'>
              {icon}
              {title} Settings
            </CardTitle>
            <CardDescription className='mt-1.5'>{description}</CardDescription>
          </div>
          <Badge variant='outline' className='ml-2'>
            {items.length} {items.length === 1 ? title.toLowerCase() : `${title.toLowerCase()}s`}
          </Badge>
        </div>
      </CardHeader>

      <Separator />

      <CardContent className='pt-6'>
        {/* Add new entity */}
        <div className='mb-6'>
          <form
            className='flex items-center gap-2'
            onSubmit={(e) => {
              e.preventDefault()
              handleAddItem()
            }}
          >
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder={`Enter new ${title.toLowerCase()} name`}
              className='flex-1'
              disabled={createMutation.isPending}
            />
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button type='submit' disabled={!newName.trim() || createMutation.isPending}>
                    {createMutation.isPending ? (
                      <Loader2 className='h-4 w-4 mr-2 animate-spin' />
                    ) : (
                      <Plus className='h-4 w-4 mr-2' />
                    )}
                    Add {title}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Create a new {title.toLowerCase()}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </form>
          {createError && (
            <div className='mt-2 text-sm text-destructive flex items-center gap-1.5'>
              <AlertCircle className='h-4 w-4' />
              {createError}
            </div>
          )}
        </div>

        {/* List of entities */}
        {sortedItems.length === 0 ? (
          <div className='text-center py-8 text-muted-foreground'>
            {emptyIcon}
            <p>{emptyMessage}</p>
            <p className='text-sm mt-1'>{emptySubtext}</p>
          </div>
        ) : (
          <div className='space-y-4'>
            {sortedItems.map((item) => (
              <div
                key={item.id}
                className={`
                  flex items-center justify-between p-3 rounded-md border
                  ${editingItem?.id === item.id ? 'bg-muted/50 border-primary/20' : 'bg-card hover:bg-accent/5'}
                  transition-colors duration-200
                `}
              >
                {editingItem && editingItem.id === item.id ? (
                  <form
                    className='flex items-center w-full gap-2 flex-wrap'
                    onSubmit={(e) => {
                      e.preventDefault()
                      handleUpdateItem()
                    }}
                  >
                    <Input
                      value={editingItem.name}
                      onChange={(e) => setEditingItem({ ...editingItem, name: e.target.value })}
                      className='flex-1'
                      autoFocus
                      disabled={updateMutation.isPending}
                    />
                    <div className='flex items-center gap-1'>
                      <Button
                        size='sm'
                        type='submit'
                        disabled={!editingItem.name.trim() || updateMutation.isPending}
                      >
                        {updateMutation.isPending ? (
                          <Loader2 className='h-4 w-4 animate-spin' />
                        ) : (
                          <Save className='h-4 w-4' />
                        )}
                      </Button>
                      <Button
                        size='sm'
                        variant='ghost'
                        onClick={() => setEditingItem(null)}
                        disabled={updateMutation.isPending}
                      >
                        <X className='h-4 w-4' />
                      </Button>
                    </div>
                    {updateError && (
                      <div className='w-full mt-2 text-sm text-destructive flex items-center gap-1.5'>
                        <AlertCircle className='h-4 w-4' />
                        {updateError}
                      </div>
                    )}
                  </form>
                ) : (
                  <>
                    <span className='font-medium'>{item.name}</span>
                    <div className='flex items-center gap-2'>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              size='sm'
                              variant='ghost'
                              onClick={() => setEditingItem({ id: item.id, name: item.name })}
                              disabled={deleteMutation.isPending}
                            >
                              <Pencil className='h-4 w-4' />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>Rename {title.toLowerCase()}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                      <DeleteDialog
                        entityName={title.toLowerCase()}
                        name={item.name}
                        disabled={deleteMutation.isPending && deleteMutation.variables === item.id}
                        onDelete={() => handleDeleteItem(item.id)}
                      />
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}
        {deleteError && (
          <div className='mt-4 text-sm text-destructive flex items-center gap-1.5 p-2 bg-destructive/5 rounded-sm border border-destructive/20'>
            <AlertCircle className='h-4 w-4' />
            {deleteError}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
